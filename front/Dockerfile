FROM node:18-alpine

RUN npm config set registry https://registry.npmmirror.com/

WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .

EXPOSE 3000
CMD ["npm", "run", "dev"]
