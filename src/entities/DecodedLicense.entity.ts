import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, OneToOne, JoinColumn } from "typeorm";
import { License } from "./License.entity";

@Entity("decoded_licenses")
export class DecodedLicense {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column()
  company!: string;

  @Column()
  licenseType!: string;

  @Column({ type: "timestamp" })
  expiry!: Date;

  @Column({ type: "timestamp" })
  issuedAt!: Date;

  @Column({ type: "jsonb" })
  tokens!: any[];

  @Column({ default: true })
  isActive!: boolean;

  @Column({ default: false })
  isDeleted!: boolean;

  @CreateDateColumn()
  createdAt!: Date;

  @Column({ type: "uuid", nullable: true })
  licenseId!: string | null;

  @Column({ type: "varchar", nullable: true })
  softwareVersion!: string | null;

  @Column({ type: "varchar", nullable: true })
  externalLicenseId!: string | null;

  @OneToOne(() => License, (l) => l.decodedLicense, { onDelete: "CASCADE", nullable: true })
  @JoinColumn({ name: "licenseId" })
  licenseBundle!: License | null;
}
