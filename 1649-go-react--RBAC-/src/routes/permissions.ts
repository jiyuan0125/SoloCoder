import express from 'express';
import * as permissionService from '../services/permissionService';
import * as userService from '../services/userService';

const router = express.Router();
router.use(express.json());

router.post('/', (req, res) => {
  try {
    const { name } = req.body;
    if (!name || typeof name !== 'string') {
      return res.status(400).json({ error: 'Permission name is required and must be a string' });
    }

    const permission = permissionService.createPermission(name);
    res.status(201).json(permission);
  } catch (e: any) {
    if (e.message.includes('Invalid permission format')) {
      return res.status(400).json({ 
        error: e.message,
        format: 'resource:action (e.g., "user:read", "order:write", "product:delete")'
      });
    }
    res.status(500).json({ error: e.message });
  }
});

router.get('/', (req, res) => {
  try {
    const permissions = permissionService.getAllPermissions();
    res.json(permissions);
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.post('/check', (req, res) => {
  try {
    const { userId, resource, action } = req.body;
    
    if (!userId || typeof userId !== 'string') {
      return res.status(400).json({ error: 'userId is required and must be a string' });
    }
    if (!resource || typeof resource !== 'string') {
      return res.status(400).json({ error: 'resource is required and must be a string' });
    }
    if (!action || typeof action !== 'string') {
      return res.status(400).json({ error: 'action is required and must be a string' });
    }

    const allowed = userService.checkUserPermission(userId, resource, action);
    res.json({ allowed });
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

export default router;
