import express from 'express';
import { initDatabase } from './database';
import projectsRouter from './routes/projects';
import artifactsRouter from './routes/artifacts';
import * as fs from 'fs';
import path from 'path';

const app = express();
const PORT = process.env.PORT || 3000;

const dataDirs = [
  path.join(process.cwd(), 'data'),
  path.join(process.cwd(), 'data', 'uploads'),
  path.join(process.cwd(), 'data', 'storage')
];

for (const dir of dataDirs) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

initDatabase();

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/projects', projectsRouter);
app.use('/artifacts', artifactsRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`Artifact manager server running on port ${PORT}`);
});
