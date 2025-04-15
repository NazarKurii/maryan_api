// server.js or app.js
const express = require('express');
const swaggerUi = require('swagger-ui-express');
const YAML = require('yamljs');
const path = require('path');

const app = express();

// Load your OpenAPI YAML or JSON file
const swaggerDocument = YAML.load(path.join(__dirname, './openapidist.yaml'));

app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerDocument));

app.listen(3000, () => {
  console.log('API docs available at http://localhost:3000/api-docs');
});
