import express from 'express';
import taskGroupsRouter, { default as taskGroups } from './routes/taskGroups';
import tasksRouter, { setScheduler } from './routes/tasks';
import { initializeScheduler, scheduleTask } from './scheduler';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/task-groups', taskGroupsRouter);
app.use('/tasks', tasksRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

async function start() {
  try {
    await initializeScheduler();
    setScheduler(scheduleTask);
    
    app.listen(PORT, () => {
      console.log(`Task scheduler server running on port ${PORT}`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

start();
