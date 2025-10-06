FROM node:18-alpine

WORKDIR /app

RUN apk add --no-cache tcpdump python3 py3-pip python3-dev build-base

COPY runner/package.json runner/pnpm-lock.yaml ./
COPY runner/services/pine-parser/requirements.txt ./services/pine-parser/
COPY PineTS /PineTS
RUN npm install -g pnpm@10 && pnpm install --frozen-lockfile
RUN pip3 install --break-system-packages --no-cache-dir -r services/pine-parser/requirements.txt

CMD ["pnpm", "start"]
