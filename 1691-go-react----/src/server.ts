import express from 'express';
import fs from 'fs';
import path from 'path';
import organizationsRouter from './routes/organizations';
import inspectionsRouter from './routes/inspections';
import evaluationsRouter from './routes/evaluations';

const dataDir = path.join(__dirname, '..', 'data');
if (!fs.existsSync(dataDir)) {
  fs.mkdirSync(dataDir, { recursive: true });
}

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/', (req, res) => {
  res.json({
    name: '社会组织管理系统',
    version: '1.0.0',
    endpoints: {
      organizations: '/api/organizations',
      inspections: '/api/inspections',
      evaluations: '/api/evaluations'
    }
  });
});

app.use('/api/organizations', organizationsRouter);
app.use('/api/inspections', inspectionsRouter);
app.use('/api/evaluations', evaluationsRouter);

app.listen(PORT, () => {
  console.log(`社会组织管理系统运行在 http://localhost:${PORT}`);
});

export default app;
