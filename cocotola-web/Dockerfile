ARG PNPM_VERSION=10.33.0
ARG NODE_VERSION=22.22.2

FROM node:${NODE_VERSION}-alpine AS development-dependencies-env
ARG PNPM_VERSION
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
COPY package.json pnpm-lock.yaml /app/
WORKDIR /app
RUN pnpm install --frozen-lockfile

FROM node:${NODE_VERSION}-alpine AS production-dependencies-env
ARG PNPM_VERSION
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
COPY package.json pnpm-lock.yaml /app/
WORKDIR /app
RUN pnpm install --frozen-lockfile --prod

FROM node:${NODE_VERSION}-alpine AS build-env
ARG PNPM_VERSION
ARG VITE_SUPABASE_URL
ARG VITE_SUPABASE_ANON_KEY
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
COPY . /app/
COPY --from=development-dependencies-env /app/node_modules /app/node_modules
WORKDIR /app
RUN pnpm run build

FROM node:${NODE_VERSION}-alpine
ARG PNPM_VERSION
RUN corepack enable && corepack prepare pnpm@${PNPM_VERSION} --activate
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY package.json pnpm-lock.yaml /app/
COPY --from=production-dependencies-env /app/node_modules /app/node_modules
COPY --from=build-env /app/build /app/build
WORKDIR /app
USER appuser
CMD ["pnpm", "run", "start"]
