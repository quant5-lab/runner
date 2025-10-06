FROM node:18-alpine

WORKDIR /app

RUN apk add --no-cache tcpdump

COPY runner/package.json runner/pnpm-lock.yaml ./
COPY PineTS /PineTS
RUN npm install -g pnpm@10 && pnpm install --frozen-lockfile

CMD ["pnpm", "start"]
