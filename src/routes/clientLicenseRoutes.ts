import { Router } from 'express';
import { importLicense, getLicense, getAllLicenses, getLicenseById, toggleLicenseStatus, deleteLicense } from '../controllers/clientLicenseController';

const router = Router();


// POST /api/client-license/import
router.post('/import', importLicense);

// GET /api/client-license/all
router.get('/all', getAllLicenses);

// PATCH /api/client-license/:id/status
router.patch('/:id/status', toggleLicenseStatus);

// DELETE /api/client-license/:id
router.delete('/:id', deleteLicense);

// GET /api/client-license/:id
router.get('/:id', getLicenseById);


// GET /api/client-license/
router.get('/', getLicense);



export default router;
