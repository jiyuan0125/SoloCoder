import express from 'express';
import bodyParser from 'body-parser';
import cors from 'cors';
import { usersRouter } from './routes/users';
import { issuesRouter } from './routes/issues';
import { todosRouter } from './routes/todos';
import './database';

const app = express();
const PORT = process.env.PORT || 8110;

app.use(cors());
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));

app.use('/api/users', usersRouter);
app.use('/api/issues', issuesRouter);
app.use('/api/todos', todosRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`Community Voting System running on port ${PORT}`);
});

export { app };
