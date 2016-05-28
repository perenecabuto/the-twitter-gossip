var BrowserSyncPlugin = require('browser-sync-webpack-plugin');

module.exports = {
    node: {
        global: true
    },
    plugins: [
        new BrowserSyncPlugin({
            host: 'localhost',
            port: 3000,
            server: { baseDir: ['.'] }
        })
    ],

    module: {
        loaders: [
            {
                test: /\.jsx?$/,
                exclude: /(node_modules|bower_components)/,
                loader: 'babel', // 'babel-loader' is also a legal name to reference
                query: {
                    presets: ['es2015', 'react']
                }
            }
        ]
    },

    entry: "./main.jsx",
    output: {
        filename: "bundle.js"
    }
}
