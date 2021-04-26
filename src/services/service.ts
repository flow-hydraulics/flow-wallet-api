import {PrismaClient} from "@prisma/client"

export default class Service {
  protected prisma: PrismaClient

  constructor(prisma: PrismaClient) {
    this.prisma = prisma
  }
}
