import express from 'express';
import * as userService from '../services/userService';

const router = express.Router();
router.use(express.json());

router.post('/', (req, res) => {
  try {
    const { name, id } = req.body;
    if (!name || typeof name !== 'string') {
      return res.status(400).json({ error: 'User name is required and must be a string' });
    }

    const user = userService.createUser(name, id);
    res.status(201).json(user);
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.get('/', (req, res) => {
  try {
    const users = userService.getAllUsers();
    res.json(users);
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.get('/:userId', (req, res) => {
  try {
    const user = userService.getUserById(req.params.userId);
    if (user) {
      const roles = userService.getUserRoles(req.params.userId);
      const permissions = userService.getUserPermissions(req.params.userId);
      res.json({
        user,
        roles,
        permissions
      });
    } else {
      res.status(404).json({ error: 'User not found' });
    }
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.post('/:userId/roles', (req, res) => {
  try {
    const userId = req.params.userId;
    const { roleId } = req.body;
    
    if (!roleId || typeof roleId !== 'string') {
      return res.status(400).json({ error: 'roleId is required and must be a string' });
    }

    userService.assignRoleToUser(userId, roleId);
    res.status(204).send();
  } catch (e: any) {
    if (e.message === 'User already has this role') {
      return res.status(409).json({ 
        error: 'User already has this role',
        message: 'This role is already assigned to the user'
      });
    }
    if (e.message === 'Role not found') {
      return res.status(404).json({ error: e.message });
    }
    res.status(500).json({ error: e.message });
  }
});

router.delete('/:userId/roles/:roleId', (req, res) => {
  try {
    userService.removeRoleFromUser(req.params.userId, req.params.roleId);
    res.status(204).send();
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

export default router;
