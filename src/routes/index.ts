import { Router } from 'express';
import clientLicenseRoutes from './clientLicenseRoutes';

const router = Router();

// Health Check
router.get('/health', (_req, res) => {
  res.json({ status: 'ok', service: 'get-license-api', timestamp: new Date().toISOString() });
});

// API Routes
router.use('/api/client-license', clientLicenseRoutes);

export default router;
