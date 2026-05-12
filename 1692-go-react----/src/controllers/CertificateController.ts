import { Request, Response } from 'express';
import { CertificateService } from '../services/CertificateService';

export class CertificateController {
  private certificateService: CertificateService;

  constructor() {
    this.certificateService = new CertificateService();
  }

  issueCertificate = async (req: Request, res: Response): Promise<void> => {
    try {
      const { studentId, courseId } = req.body;
      const certificate = await this.certificateService.issueCertificate(studentId, courseId);
      res.status(201).json(certificate);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '颁发证书失败'
      });
    }
  };

  getCertificate = async (req: Request, res: Response): Promise<void> => {
    try {
      const { certificateId } = req.params;
      const certificate = await this.certificateService.getCertificateById(certificateId);
      res.status(200).json(certificate);
    } catch (error: any) {
      const statusCode = error.statusCode || 500;
      res.status(statusCode).json({
        success: false,
        message: error.message || '获取证书失败'
      });
    }
  };

  getStudentCertificates = async (req: Request, res: Response): Promise<void> => {
    try {
      const { studentId } = req.params;
      const certificates = await this.certificateService.getStudentCertificates(studentId);
      res.status(200).json(certificates);
    } catch (error: any) {
      res.status(500).json({
        success: false,
        message: error.message || '获取学员证书失败'
      });
    }
  };
}
