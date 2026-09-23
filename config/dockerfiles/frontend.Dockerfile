# Frontend de desarrollo: Node 24 LTS + el proyecto Angular. El código entra por
# volumen desde la máquina (ver dev.yml); el node_modules vive en un volumen con
# nombre, que Docker rellena con lo que instala esta imagen.
FROM node:24-alpine

WORKDIR /app

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./

EXPOSE 11001

CMD ["npm", "start", "--", "--host", "0.0.0.0", "--port", "11001"]
