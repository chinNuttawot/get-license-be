import { Request, Response } from 'express';
import crypto from 'crypto';
import { catchAsync } from '../utils/catchAsync';
import { AppError } from '../utils/AppError';
import { AppDataSource } from '../config/database.config';
import { License } from '../entities/License.entity';
import { DecodedLicense } from '../entities/DecodedLicense.entity';

function loadPublicKey(raw: string): crypto.KeyObject {
  const cleaned = raw.replace(/\\n/g, '').replace(/\s/g, '');
  const body = (cleaned.match(/.{1,64}/g) || []).join('\n');
  const pem = `-----BEGIN PUBLIC KEY-----\n${body}\n-----END PUBLIC KEY-----\n`;
  return crypto.createPublicKey({ key: pem, format: 'pem', type: 'spki' });
}

export const importLicense = catchAsync(async (req: Request, res: Response) => {
  const { fileContent } = req.body;
  if (!fileContent) {
    throw new AppError(400, 'fileContent is required. Please provide the base64 string of the .aglic file.');
  }

  const BUNDLE_KEY = process.env.BUNDLE_KEY;
  if (!BUNDLE_KEY) {
    throw new AppError(500, 'BUNDLE_KEY is not configured in the environment.');
  }

  const key = Buffer.from(BUNDLE_KEY, 'base64');

  let envelope;
  try {
    const envelopeJson = Buffer.from(fileContent, 'base64').toString('utf8');
    envelope = JSON.parse(envelopeJson);
  } catch (err) {
    throw new AppError(400, 'Invalid file content. Could not decode base64 or parse JSON.');
  }

  if (envelope.v !== 1 || envelope.alg !== 'AES-GCM-256') {
    throw new AppError(400, 'Unsupported bundle format or algorithm.');
  }

  const iv = Buffer.from(envelope.iv, 'base64');
  const tag = Buffer.from(envelope.tag, 'base64');
  const ciphertext = Buffer.from(envelope.data, 'base64');

  const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv);
  decipher.setAuthTag(tag);

  let bundle;
  try {
    const decrypted = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
    bundle = JSON.parse(decrypted.toString());
  } catch (err) {
    throw new AppError(400, 'Decryption failed. BUNDLE_KEY is likely incorrect or data is corrupted.');
  }

  const queryRunner = AppDataSource.createQueryRunner();
  await queryRunner.connect();
  await queryRunner.startTransaction();

  try {
    // 1. Save License (Table 1)
    const rawRepo = queryRunner.manager.getRepository(License);
    const rawLicense = rawRepo.create({
      content: fileContent
    });
    await rawRepo.save(rawLicense);

    // 2. Save Decoded License (Table 2)
    const decodedRepo = queryRunner.manager.getRepository(DecodedLicense);
    const decodedLicense = decodedRepo.create({
      licenseId: rawLicense.id,
      company: bundle.meta.company,
      licenseType: bundle.meta.licenseType,
      expiry: new Date(bundle.meta.expiry),
      issuedAt: new Date(bundle.meta.issuedAt),
      tokens: bundle.tokens
    });
    await decodedRepo.save(decodedLicense);

    await queryRunner.commitTransaction();

    res.status(201).json({
      message: 'License imported successfully',
      data: {
        rawId: rawLicense.id,
        decodedId: decodedLicense.id,
        meta: bundle.meta
      }
    });
  } catch (err) {
    await queryRunner.rollbackTransaction();
    throw new AppError(500, 'Failed to save license to database.');
  } finally {
    await queryRunner.release();
  }
});

export const getLicense = catchAsync(async (req: Request, res: Response) => {
  const decodedRepo = AppDataSource.getRepository(DecodedLicense);
  
  // Get the latest imported license
  const latestLicense = await decodedRepo.findOne({
    where: {},
    order: { createdAt: 'DESC' }
  });

  if (!latestLicense) {
    return res.status(404).json({ message: 'No license found.' });
  }

  const PUBLIC_KEY_RAW = process.env.PUBLIC_KEY;
  if (!PUBLIC_KEY_RAW) {
    throw new AppError(500, 'PUBLIC_KEY is not configured in the environment.');
  }

  let publicKey: crypto.KeyObject;
  try {
    publicKey = loadPublicKey(PUBLIC_KEY_RAW);
  } catch (err) {
    throw new AppError(500, 'Failed to load PUBLIC_KEY.');
  }

  const verifiedTokens = latestLicense.tokens.map((t: string, i: number) => {
    try {
      const decoded = JSON.parse(Buffer.from(t, 'base64').toString('utf8'));
      const isVerified = crypto.verify(
        undefined, 
        Buffer.from(decoded.data),
        publicKey,
        Buffer.from(decoded.signature, 'hex')
      );

      return {
        index: i + 1,
        isVerified,
        data: isVerified ? JSON.parse(decoded.data) : null,
        error: isVerified ? null : 'INVALID_SIGNATURE'
      };
    } catch (err) {
      return {
        index: i + 1,
        isVerified: false,
        data: null,
        error: (err as Error).message
      };
    }
  });

  res.json({
    meta: {
      company: latestLicense.company,
      licenseType: latestLicense.licenseType,
      expiry: latestLicense.expiry,
      issuedAt: latestLicense.issuedAt,
      importedAt: latestLicense.createdAt
    },
    tokens: verifiedTokens
  });
});

export const getAllLicenses = catchAsync(async (req: Request, res: Response) => {
  const decodedRepo = AppDataSource.getRepository(DecodedLicense);
  
  const allLicenses = await decodedRepo.find({
    where: { isDeleted: false }, // Filter out deleted
    order: { createdAt: 'DESC' }
  });

  res.json({
    data: allLicenses.map(l => ({
      id: l.id,
      company: l.company,
      licenseType: l.licenseType,
      expiry: l.expiry,
      issuedAt: l.issuedAt,
      importedAt: l.createdAt,
      tokenCount: l.tokens.length,
      isActive: l.isActive
    }))
  });
});

export const getLicenseById = catchAsync(async (req: Request, res: Response) => {
  const { id } = req.params;
  const decodedRepo = AppDataSource.getRepository(DecodedLicense);
  
  const license = await decodedRepo.findOne({
    where: { id, isDeleted: false }
  });

  if (!license) {
    throw new AppError(404, 'License not found.');
  }

  const PUBLIC_KEY_RAW = process.env.PUBLIC_KEY;
  if (!PUBLIC_KEY_RAW) {
    throw new AppError(500, 'PUBLIC_KEY is not configured in the environment.');
  }

  let publicKey: crypto.KeyObject;
  try {
    publicKey = loadPublicKey(PUBLIC_KEY_RAW);
  } catch (err) {
    throw new AppError(500, 'Failed to load PUBLIC_KEY.');
  }

  const verifiedTokens = license.tokens.map((t: string, i: number) => {
    try {
      const decoded = JSON.parse(Buffer.from(t, 'base64').toString('utf8'));
      const isVerified = crypto.verify(
        undefined, 
        Buffer.from(decoded.data),
        publicKey,
        Buffer.from(decoded.signature, 'hex')
      );

      return {
        index: i + 1,
        isVerified: isVerified && license.isActive, // If license inactive, tokens are invalid
        data: (isVerified && license.isActive) ? JSON.parse(decoded.data) : null,
        error: !license.isActive ? 'LICENSE_INACTIVE' : (isVerified ? null : 'INVALID_SIGNATURE')
      };
    } catch (err) {
      return {
        index: i + 1,
        isVerified: false,
        data: null,
        error: (err as Error).message
      };
    }
  });

  res.json({
    meta: {
      company: license.company,
      licenseType: license.licenseType,
      expiry: license.expiry,
      issuedAt: license.issuedAt,
      importedAt: license.createdAt,
      isActive: license.isActive
    },
    tokens: verifiedTokens
  });
});

export const toggleLicenseStatus = catchAsync(async (req: Request, res: Response) => {
  const { id } = req.params;
  const decodedRepo = AppDataSource.getRepository(DecodedLicense);
  
  const license = await decodedRepo.findOne({ where: { id, isDeleted: false } });
  if (!license) throw new AppError(404, 'License not found');

  license.isActive = !license.isActive;
  await decodedRepo.save(license);

  res.json({ message: `License ${license.isActive ? 'activated' : 'deactivated'}`, data: { isActive: license.isActive } });
});

export const deleteLicense = catchAsync(async (req: Request, res: Response) => {
  const { id } = req.params;
  const decodedRepo = AppDataSource.getRepository(DecodedLicense);
  
  const license = await decodedRepo.findOne({ where: { id } });
  if (!license) throw new AppError(404, 'License not found');

  license.isDeleted = true;
  await decodedRepo.save(license);

  res.json({ message: 'License deleted successfully' });
});


