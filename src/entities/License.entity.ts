import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, OneToOne } from "typeorm";
import { DecodedLicense } from "./DecodedLicense.entity";

@Entity("licenses")
export class License {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "text" })
  content!: string;

  @CreateDateColumn()
  createdAt!: Date;

  @OneToOne(() => DecodedLicense, (decoded) => decoded.licenseBundle)
  decodedLicense!: DecodedLicense;
}
