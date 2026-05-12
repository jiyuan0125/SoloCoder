import { Router } from 'express';
import { CertificateController } from '../controllers/CertificateController';

export class CertificateRoutes {
  public router: Router;
  private certificateController: CertificateController;

  constructor() {
    this.router = Router();
    this.certificateController = new CertificateController();
    this.initializeRoutes();
  }

  private initializeRoutes(): void {
    this.router.post('/', this.certificateController.issueCertificate);
    this.router.get('/:certificateId', this.certificateController.getCertificate);
    this.router.get('/student/:studentId', this.certificateController.getStudentCertificates);
  }
}
