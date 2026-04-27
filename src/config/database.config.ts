import { DataSource } from "typeorm";
import * as dotenv from "dotenv";
import * as path from "path";

dotenv.config({ path: path.join(__dirname, "../../.env") });

export const AppDataSource = new DataSource({
  type: "postgres",
  host: process.env.DB_HOST,
  port: parseInt(process.env.DB_PORT || "5432"),
  username: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME,
  synchronize: true, // Auto schema sync
  logging: false,
  entities: [
    path.join(__dirname, "../entities/**/*.entity{.ts,.js}")
  ],
  subscribers: [],
  migrations: [],
  ssl: process.env.DB_SSL !== "false",
  extra: process.env.DB_SSL !== "false" ? {
    ssl: {
      rejectUnauthorized: false
    }
  } : undefined
});

export const connectDB = async () => {
  try {
    await AppDataSource.initialize();
    console.log("📦 Data Source has been initialized!");
  } catch (err) {
    console.error("❌ Error during Data Source initialization:", err);
    process.exit(1);
  }
};
