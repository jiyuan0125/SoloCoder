import express from 'express';
import * as roleService from '../services/roleService';
import * as permissionService from '../services/permissionService';

const router = express.Router();
router.use(express.json());

router.post('/', (req, res) => {
  try {
    const { name, description } = req.body;
    if (!name || typeof name !== 'string') {
      return res.status(400).json({ error: 'Role name is required and must be a string' });
    }

    const role = roleService.createRole(name, description);
    res.status(201).json(role);
  } catch (e: any) {
    if (e.message === 'Role with this name already exists') {
      return res.status(409).json({ error: e.message });
    }
    res.status(500).json({ error: e.message });
  }
});

router.get('/', (req, res) => {
  try {
    const roles = roleService.getAllRoles();
    res.json(roles);
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.get('/:roleId', (req, res) => {
  try {
    const role = roleService.getRoleWithPermissions(req.params.roleId);
    if (role) {
      res.json(role);
    } else {
      res.status(404).json({ error: 'Role not found' });
    }
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

router.delete('/:roleId', (req, res) => {
  try {
    const roleId = req.params.roleId;
    roleService.deleteRole(roleId);
    res.status(204).send();
  } catch (e: any) {
    if (e.message.startsWith('Role is still assigned')) {
      const match = e.message.match(/users: (.+)$/);
      const userIds = match ? match[1].split(', ') : [];
      return res.status(409).json({ 
        error: e.message,
        users: userIds,
        message: 'Cannot delete role because it is still assigned to users'
      });
    }
    if (e.message === 'Role not found') {
      return res.status(404).json({ error: e.message });
    }
    res.status(500).json({ error: e.message });
  }
});

router.post('/:roleId/permissions', (req, res) => {
  try {
    const roleId = req.params.roleId;
    const { permissionId, permissionName } = req.body;
    
    let actualPermissionId = permissionId;
    
    if (!actualPermissionId && permissionName) {
      const perm = permissionService.getPermissionByName(permissionName);
      if (!perm) {
        return res.status(404).json({ error: 'Permission not found' });
      }
      actualPermissionId = perm.id;
    }
    
    if (!actualPermissionId) {
      return res.status(400).json({ error: 'permissionId or permissionName is required' });
    }

    roleService.assignPermissionToRole(roleId, actualPermissionId);
    res.status(204).send();
  } catch (e: any) {
    if (e.message === 'Role not found' || e.message === 'Permission not found') {
      return res.status(404).json({ error: e.message });
    }
    res.status(500).json({ error: e.message });
  }
});

router.post('/:roleId/inherit/:parentRoleId', (req, res) => {
  try {
    const childRoleId = req.params.roleId;
    const parentRoleId = req.params.parentRoleId;

    roleService.setRoleInheritance(childRoleId, parentRoleId);
    res.status(204).send();
  } catch (e: any) {
    if (e.message === 'Role not found') {
      return res.status(404).json({ error: e.message });
    }
    if (e.message === 'Cannot inherit from self') {
      return res.status(400).json({ error: e.message });
    }
    if (e.message === 'Inheritance would create a cycle') {
      return res.status(400).json({ 
        error: e.message,
        reason: 'Setting this inheritance would create a cycle in the role hierarchy'
      });
    }
    res.status(500).json({ error: e.message });
  }
});

export default router;
