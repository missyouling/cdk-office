# CDK-Office Frontend Dockerfile

# Build stage
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install pnpm globally
RUN npm install -g pnpm

# Install dependencies
RUN pnpm install

# Copy source code
COPY . .

# Build the application
RUN pnpm build

# Production stage
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install pnpm globally
RUN npm install -g pnpm

# Install production dependencies only
RUN pnpm install --prod

# Copy built files from builder stage
COPY --from=builder /app/.next ./.next

# Expose port
EXPOSE 3000

# Start the application
CMD ["pnpm", "start"]