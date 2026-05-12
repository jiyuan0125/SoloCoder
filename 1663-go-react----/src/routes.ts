import { Router, Request, Response } from 'express';
import { channelService } from './services/channelService';
import { linkService } from './services/linkService';
import { attributionService } from './services/attributionService';
import { orderCommissionService } from './services/orderCommissionService';
import { CreateChannelRequest, UpdateChannelRequest, CreateLinkRequest, TrackAttributionRequest, CreateOrderRequest } from './types';

const router = Router();

// 渠道管理路由
router.post('/channels', (req: Request, res: Response) => {
  try {
    const data = req.body as CreateChannelRequest;
    const channel = channelService.createChannel(data);
    res.status(201).json(channel);
  } catch (error) {
    if (error instanceof Error) {
      if (error.message === '佣金比例必须在 0 到 1 之间') {
        res.status(400).json({ error: error.message });
      } else if (error.message === '渠道名称已存在') {
        res.status(409).json({ error: error.message });
      } else {
        res.status(500).json({ error: error.message });
      }
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.get('/channels', (req: Request, res: Response) => {
  try {
    const channels = channelService.getAllChannels();
    res.json(channels);
  } catch (error) {
    res.status(500).json({ error: '获取渠道列表失败' });
  }
});

router.get('/channels/:id', (req: Request, res: Response) => {
  try {
    const channel = channelService.getChannelById(req.params.id);
    if (!channel) {
      return res.status(404).json({ error: '渠道不存在' });
    }
    res.json(channel);
  } catch (error) {
    res.status(500).json({ error: '获取渠道信息失败' });
  }
});

router.put('/channels/:id', (req: Request, res: Response) => {
  try {
    const data = req.body as UpdateChannelRequest;
    const channel = channelService.updateChannel(req.params.id, data);
    res.json(channel);
  } catch (error) {
    if (error instanceof Error) {
      if (error.message === '渠道不存在') {
        res.status(404).json({ error: error.message });
      } else if (error.message === '佣金比例必须在 0 到 1 之间') {
        res.status(400).json({ error: error.message });
      } else if (error.message === '渠道名称已存在') {
        res.status(409).json({ error: error.message });
      } else {
        res.status(500).json({ error: error.message });
      }
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.delete('/channels/:id', (req: Request, res: Response) => {
  try {
    channelService.deleteChannel(req.params.id);
    res.status(204).send();
  } catch (error) {
    if (error instanceof Error) {
      res.status(404).json({ error: error.message });
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

// 推广链接管理路由
router.post('/links', (req: Request, res: Response) => {
  try {
    const data = req.body as CreateLinkRequest;
    const link = linkService.createLink(data);
    res.status(201).json(link);
  } catch (error) {
    if (error instanceof Error) {
      res.status(400).json({ error: error.message });
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.get('/links', (req: Request, res: Response) => {
  try {
    const links = linkService.getAllLinks();
    res.json(links);
  } catch (error) {
    res.status(500).json({ error: '获取链接列表失败' });
  }
});

router.get('/links/:id', (req: Request, res: Response) => {
  try {
    const link = linkService.getLinkById(req.params.id);
    if (!link) {
      return res.status(404).json({ error: '链接不存在' });
    }
    res.json(link);
  } catch (error) {
    res.status(500).json({ error: '获取链接信息失败' });
  }
});

router.get('/channels/:channelId/links', (req: Request, res: Response) => {
  try {
    const links = linkService.getLinksByChannelId(req.params.channelId);
    res.json(links);
  } catch (error) {
    res.status(500).json({ error: '获取渠道链接列表失败' });
  }
});

// 归因追踪路由
router.post('/attributions', (req: Request, res: Response) => {
  try {
    const data = req.body as TrackAttributionRequest;
    const attribution = attributionService.trackAttribution(data);
    res.status(201).json(attribution);
  } catch (error) {
    if (error instanceof Error) {
      if (error.message === '用户ID和链接ID不能为空') {
        res.status(400).json({ error: error.message });
      } else {
        res.status(500).json({ error: error.message });
      }
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.get('/attributions/:userId', (req: Request, res: Response) => {
  try {
    const attributions = attributionService.getAttributionsByUserId(req.params.userId);
    res.json(attributions);
  } catch (error) {
    res.status(500).json({ error: '获取用户归因记录失败' });
  }
});

// 订单和佣金管理路由
router.post('/orders', (req: Request, res: Response) => {
  try {
    const data = req.body as CreateOrderRequest;
    const result = orderCommissionService.createOrder(data);
    res.status(201).json(result);
  } catch (error) {
    if (error instanceof Error) {
      if (error.message === '用户ID不能为空' || error.message === '订单金额必须大于0') {
        res.status(400).json({ error: error.message });
      } else if (error.message === '订单已归因') {
        res.status(409).json({ error: error.message });
      } else {
        res.status(500).json({ error: error.message });
      }
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.post('/orders/:orderId/refund', (req: Request, res: Response) => {
  try {
    orderCommissionService.refundOrder(req.params.orderId);
    res.status(200).json({ message: '订单已退款' });
  } catch (error) {
    if (error instanceof Error) {
      if (error.message === '订单不存在') {
        res.status(404).json({ error: error.message });
      } else if (error.message === '已结算的佣金需要先走退款流程') {
        res.status(400).json({ error: error.message });
      } else {
        res.status(500).json({ error: error.message });
      }
    } else {
      res.status(500).json({ error: '服务器错误' });
    }
  }
});

router.get('/orders/:id', (req: Request, res: Response) => {
  try {
    const order = orderCommissionService.getOrderById(req.params.id);
    if (!order) {
      return res.status(404).json({ error: '订单不存在' });
    }
    res.json(order);
  } catch (error) {
    res.status(500).json({ error: '获取订单信息失败' });
  }
});

router.get('/commissions/:id', (req: Request, res: Response) => {
  try {
    const commission = orderCommissionService.getCommissionById(req.params.id);
    if (!commission) {
      return res.status(404).json({ error: '佣金记录不存在' });
    }
    res.json(commission);
  } catch (error) {
    res.status(500).json({ error: '获取佣金信息失败' });
  }
});

router.get('/channels/:channelId/commissions', (req: Request, res: Response) => {
  try {
    const commissions = orderCommissionService.getCommissionsByChannelId(req.params.channelId);
    res.json(commissions);
  } catch (error) {
    res.status(500).json({ error: '获取渠道佣金列表失败' });
  }
});

router.get('/settlement-stats', (req: Request, res: Response) => {
  try {
    const stats = orderCommissionService.getSettlementStats();
    res.json(stats);
  } catch (error) {
    res.status(500).json({ error: '获取结算统计失败' });
  }
});

router.post('/settle', (req: Request, res: Response) => {
  try {
    const { channelId, settlementMonth } = req.body;
    orderCommissionService.settleCommissions(channelId, settlementMonth);
    res.status(200).json({ message: '结算完成' });
  } catch (error) {
    res.status(500).json({ error: '结算失败' });
  }
});

export default router;
