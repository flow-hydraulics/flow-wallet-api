FROM node:14-alpine3.13

WORKDIR /app

COPY package.json .
COPY src/database/schema.prisma src/database/schema.prisma

RUN npm install

ADD . /app

RUN npm run build

CMD [ "npm", "start" ]

EXPOSE 3000
