import app from './app';

const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.listen(PORT, () => {
  console.log(`Anti-Spam System running on port ${PORT}`);
});