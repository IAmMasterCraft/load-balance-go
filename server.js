// script to start multiple instances of the express server on different ports
// each ports returns a json message with the port number

const express = require('express');

const portList = [
    8081, 8082, 8083, 8084, 8085
];

portList.forEach(port => {
    const app = express();
    
    app.listen(port, () => {
        console.log(`Server running on port ${port}`);
    });

    app.get('/', (req, res) => {
        res.json({
            message: `Server running on port ${port}`
        });
        console.log(port);
    });
});