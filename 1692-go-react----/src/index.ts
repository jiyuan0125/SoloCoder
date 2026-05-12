import 'reflect-metadata';
import express from 'express';
import cors from 'cors';
import { initializeDatabase } from './config/database';
import { CourseRoutes } from './routes/courseRoutes';
import { InstructorRoutes } from './routes/instructorRoutes';
import { StudentRoutes } from './routes/studentRoutes';
import { CertificateRoutes } from './routes/certificateRoutes';

class App {
  public app: express.Application;
  public port: number;

  constructor() {
    this.app = express();
    this.port = parseInt(process.env.PORT || '8202', 10);
    this.initializeMiddlewares();
    this.initializeRoutes();
  }

  private initializeMiddlewares(): void {
    this.app.use(cors());
    this.app.use(express.json());
    this.app.use(express.urlencoded({ extended: true }));
  }

  private initializeRoutes(): void {
    const courseRoutes = new CourseRoutes();
    const instructorRoutes = new InstructorRoutes();
    const studentRoutes = new StudentRoutes();
    const certificateRoutes = new CertificateRoutes();

    this.app.use('/api/courses', courseRoutes.router);
    this.app.use('/api/instructors', instructorRoutes.router);
    this.app.use('/api/students', studentRoutes.router);
    this.app.use('/api/certificates', certificateRoutes.router);

    this.app.get('/health', (req, res) => {
      res.status(200).json({
        status: 'ok',
        message: '公益培训平台服务运行正常'
      });
    });
  }

  public async start(): Promise<void> {
    try {
      await initializeDatabase();
      console.log('数据库连接成功');

      this.app.listen(this.port, () => {
        console.log(`公益培训平台服务已启动，监听端口 ${this.port}`);
      });
    } catch (error) {
      console.error('启动服务失败:', error);
      process.exit(1);
    }
  }
}

const app = new App();
app.start();

export default app;
