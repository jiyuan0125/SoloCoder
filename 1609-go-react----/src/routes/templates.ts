import { Router, Request, Response } from "express";
import { MessageTemplateService } from "../services/MessageTemplateService";

const router = Router();
const service = new MessageTemplateService();

router.post("/", async (req: Request, res: Response) => {
  const template = await service.create(req.body);
  if (!template) {
    res.status(400).json({ error: "Invalid template variables. Use {{variableName}} format." });
    return;
  }
  res.status(201).json(template);
});

router.get("/", async (_req: Request, res: Response) => {
  const templates = await service.findAll();
  res.json(templates);
});

router.get("/:id", async (req: Request, res: Response) => {
  const template = await service.findById(req.params.id);
  if (!template) {
    res.status(404).json({ error: "Template not found" });
    return;
  }
  res.json(template);
});

router.put("/:id", async (req: Request, res: Response) => {
  const template = await service.update(req.params.id, req.body);
  if (!template) {
    res.status(400).json({ error: "Template not found or invalid variables" });
    return;
  }
  res.json(template);
});

router.delete("/:id", async (req: Request, res: Response) => {
  const result = await service.delete(req.params.id);
  if (result.hasReferences) {
    res.status(400).json({ error: "Cannot delete template: messages still reference it" });
    return;
  }
  if (!result.success) {
    res.status(404).json({ error: "Template not found" });
    return;
  }
  res.status(204).send();
});

export default router;
