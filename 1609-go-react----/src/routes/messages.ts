import { Router, Request, Response } from "express";
import { MessageService } from "../services/MessageService";
import { TODO_ACTIONS } from "../entities/types";

const router = Router();
const service = new MessageService();

router.post("/send", async (req: Request, res: Response) => {
  const message = await service.sendMessage(req.body);
  if (!message) {
    res.status(404).json({ error: "Template not found" });
    return;
  }
  res.status(201).json(message);
});

router.get("/", async (req: Request, res: Response) => {
  const recipientId = req.query.recipientId as string;
  if (!recipientId) {
    res.status(400).json({ error: "recipientId is required" });
    return;
  }

  const filter = (req.query.filter as "all" | "unread" | "read") || "all";
  const messages = await service.getMessages(recipientId, filter);
  const unreadCount = await service.getUnreadCount(recipientId);

  res.json({
    unreadCount,
    messages,
  });
});

router.get("/:id", async (req: Request, res: Response) => {
  const recipientId = req.query.recipientId as string;
  if (!recipientId) {
    res.status(400).json({ error: "recipientId is required" });
    return;
  }

  const message = await service.findById(recipientId, req.params.id);
  if (!message) {
    res.status(404).json({ error: "Message not found" });
    return;
  }
  res.json(message);
});

router.post("/mark-read", async (req: Request, res: Response) => {
  const { recipientId, messageIds } = req.body;
  if (!recipientId || !Array.isArray(messageIds)) {
    res.status(400).json({ error: "recipientId and messageIds array are required" });
    return;
  }

  await service.markAsRead(recipientId, messageIds);
  res.json({ success: true });
});

router.post("/mark-all-read", async (req: Request, res: Response) => {
  const { recipientId } = req.body;
  if (!recipientId) {
    res.status(400).json({ error: "recipientId is required" });
    return;
  }

  await service.markAllAsRead(recipientId);
  res.json({ success: true });
});

router.delete("/:id", async (req: Request, res: Response) => {
  const { recipientId } = req.body;
  if (!recipientId) {
    res.status(400).json({ error: "recipientId is required" });
    return;
  }

  const success = await service.deleteMessage(recipientId, req.params.id);
  if (!success) {
    res.status(404).json({ error: "Message not found" });
    return;
  }
  res.status(204).send();
});

router.post("/:id/actions/:action", async (req: Request, res: Response) => {
  const { recipientId, operatorId } = req.body;
  const { id, action } = req.params;

  if (!recipientId) {
    res.status(400).json({ error: "recipientId is required" });
    return;
  }

  if (!TODO_ACTIONS.includes(action)) {
    res.status(400).json({ error: `Invalid action. Valid actions: ${TODO_ACTIONS.join(", ")}` });
    return;
  }

  const result = await service.executeAction(
    recipientId,
    id,
    action,
    operatorId || recipientId
  );

  if (!result.success) {
    res.status(result.errorCode || 400).json({ error: result.errorMessage });
    return;
  }

  res.json({ success: true });
});

export default router;
