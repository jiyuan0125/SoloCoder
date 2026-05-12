import { Router, Request, Response } from "express";
import { MessageTypeService } from "../services/MessageTypeService";

const router = Router();
const service = new MessageTypeService();

router.post("/", async (req: Request, res: Response) => {
  try {
    const type = await service.create(req.body);
    res.status(201).json(type);
  } catch (error) {
    res.status(400).json({ error: "Invalid data" });
  }
});

router.get("/", async (_req: Request, res: Response) => {
  const types = await service.findAll();
  res.json(types);
});

router.get("/:id", async (req: Request, res: Response) => {
  const type = await service.findById(req.params.id);
  if (!type) {
    res.status(404).json({ error: "Message type not found" });
    return;
  }
  res.json(type);
});

router.put("/:id", async (req: Request, res: Response) => {
  const type = await service.update(req.params.id, req.body);
  if (!type) {
    res.status(404).json({ error: "Message type not found" });
    return;
  }
  res.json(type);
});

router.delete("/:id", async (req: Request, res: Response) => {
  const success = await service.delete(req.params.id);
  if (!success) {
    res.status(404).json({ error: "Message type not found" });
    return;
  }
  res.status(204).send();
});

export default router;
