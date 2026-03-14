FROM --platform=linux/arm64 nginx:stable-alpine

# Copy built assets from build stage
COPY dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]