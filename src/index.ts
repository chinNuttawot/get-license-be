import "reflect-metadata";
import express from "express";
import cors from "cors";
import dotenv from "dotenv";
import { connectDB } from "./config/database.config";
import routes from "./routes";

dotenv.config();

const app = express();
const PORT = process.env.PORT || 4001;

// Middleware
app.use(cors());
app.use(express.json({ limit: "10mb" })); // Increase limit for base64 files
app.use(express.urlencoded({ extended: true, limit: "10mb" }));

// Routes
app.use(routes);

// Global Error Handler
app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(err.statusCode || 500).json({
    status: err.status || "error",
    message: err.message || "Internal Server Error"
  });
});

// Start Server
const startServer = async () => {
  await connectDB();
  app.listen(PORT, () => {
    console.log(`🚀 Get-License Server running on http://localhost:${PORT}`);
  });
};

startServer();
