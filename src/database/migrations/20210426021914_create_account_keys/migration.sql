-- CreateTable
CREATE TABLE "account_keys" (
    "account_address" TEXT NOT NULL,
    "index" INTEGER NOT NULL,
    "type" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL
);

-- CreateIndex
CREATE UNIQUE INDEX "account_keys.account_address_index_unique" ON "account_keys"("account_address", "index");

-- AddForeignKey
ALTER TABLE "account_keys" ADD FOREIGN KEY ("account_address") REFERENCES "accounts"("address") ON DELETE CASCADE ON UPDATE CASCADE;
