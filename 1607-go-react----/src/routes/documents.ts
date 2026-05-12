import { Router, Request, Response } from 'express';
import {
  createDocument,
  getDocument,
  updateDocument,
  deleteDocument,
  batchImportDocuments
} from '../services/documentService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { title, content, id } = req.body;
  
  if (!title || !content) {
    return res.status(400).json({
      success: false,
      message: '标题和内容不能为空'
    });
  }
  
  const result = createDocument({ id, title, content });
  
  if (!result) {
    return res.status(400).json({
      success: false,
      message: '分词结果为空，文档未创建（可能全是特殊字符或停用词）'
    });
  }
  
  res.status(201).json({
    success: true,
    data: result
  });
});

router.post('/batch', (req: Request, res: Response) => {
  const { documents } = req.body;
  
  if (!Array.isArray(documents)) {
    return res.status(400).json({
      success: false,
      message: 'documents 必须是数组'
    });
  }
  
  const result = batchImportDocuments(documents);
  
  const response: any = {
    success: true,
    successCount: result.successCount,
    totalCount: documents.length
  };
  
  if (result.skippedIds.length > 0) {
    response.warning = {
      message: '部分文档因分词结果为空已跳过',
      skippedIds: result.skippedIds
    };
  }
  
  res.json(response);
});

router.get('/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const doc = getDocument(id);
  
  if (!doc) {
    return res.status(404).json({
      success: false,
      message: '文档不存在'
    });
  }
  
  res.json({
    success: true,
    data: doc
  });
});

router.put('/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const { title, content } = req.body;
  
  const doc = getDocument(id);
  
  if (!doc) {
    return res.status(404).json({
      success: false,
      message: '文档不存在'
    });
  }
  
  const result = updateDocument(id, { title, content });
  
  if (!result) {
    return res.status(400).json({
      success: false,
      message: '分词结果为空，文档未更新（可能全是特殊字符或停用词）'
    });
  }
  
  res.json({
    success: true,
    data: result
  });
});

router.delete('/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const success = deleteDocument(id);
  
  if (!success) {
    return res.status(404).json({
      success: false,
      message: '文档不存在'
    });
  }
  
  res.json({
    success: true,
    message: '删除成功'
  });
});

export default router;
