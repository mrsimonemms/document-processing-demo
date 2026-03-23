FROM node:lts
USER node
WORKDIR /home/node/app
ENV HOST 0.0.0.0
ENV PORT 5173
COPY . .
CMD [ "npm", "run", "dev" ]
