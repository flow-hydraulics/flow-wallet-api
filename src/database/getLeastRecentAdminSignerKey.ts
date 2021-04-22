import {PrismaClient} from "@prisma/client"

const prisma = new PrismaClient()

const getLeastRecentAccountKeySql = `
UPDATE admin_signer_keys
SET updated_at = current_timestamp
WHERE index = (
  SELECT index
  FROM admin_signer_keys
  ORDER BY updated_at
  LIMIT 1
)
RETURNING index
`

export default async function getLeastRecentAdminSignerKey(): Promise<number> {
  const results = await prisma.$queryRaw(getLeastRecentAccountKeySql)
  return results[0].index
}
