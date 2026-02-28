# Simple Nginx server to serve pre-built artifacts
FROM nginx:alpine

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy pre-built assets from the host machine's dist folder
# Ensure you have run 'ng build --configuration production' locally first!
COPY dist/frontend/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]