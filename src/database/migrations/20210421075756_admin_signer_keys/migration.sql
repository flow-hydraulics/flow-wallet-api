/*
  Warnings:

  - You are about to drop the `account_keys` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropTable
DROP TABLE "account_keys";

-- CreateTable
CREATE TABLE "admin_signer_keys" (
    "index" INTEGER NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    PRIMARY KEY ("index")
);
