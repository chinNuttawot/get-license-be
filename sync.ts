import { connectDB } from "./src/config/database.config";

async function run() {
  await connectDB();
  console.log("✅ Database synchronized and tables created successfully.");
  process.exit(0);
}

run();
