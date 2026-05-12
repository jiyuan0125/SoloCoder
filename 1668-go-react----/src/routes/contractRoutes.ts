import { Router, Request, Response } from 'express';
import {
  findContractById,
  insertContract,
  getAllContracts
} from '../repositories/contractRepository';
import { Contract, ContractStatus } from '../types';
import { generateId } from '../utils/id';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const body = req.body as { code: string; name: string; status?: ContractStatus };
    
    if (!body.code || !body.name) {
      res.status(400).json({ error: '合同代码和名称不能为空' });
      return;
    }

    const contract: Contract = {
      id: generateId(),
      code: body.code,
      name: body.name,
      status: body.status || ContractStatus.APPROVED,
      createdAt: new Date().toISOString()
    };

    insertContract(contract);
    res.status(201).json(contract);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const contract = findContractById(req.params.id);
    if (!contract) {
      res.status(404).json({ error: '合同不存在' });
    } else {
      res.json(contract);
    }
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const contracts = getAllContracts();
    res.json(contracts);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

export default router;
