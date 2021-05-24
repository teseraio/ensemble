// next.config.js

const withTM = require('next-transpile-modules')(['@teseraio/tesera-oss']);

module.exports = withTM({
    pageExtensions: ['js', 'jsx', 'mdx'],

    webpack: (config, { isServer }) => {
        if (!isServer) {
            config.node = {
                fs: "empty",
                net: 'empty'
            }
        }
        config.module.rules.push({
            test: /\.svg$/,
            use: [
                {
                    loader: "@svgr/webpack",
                    options: {
                        svgo: false, // Optimization caused bugs with some of my SVGs
                    },
                },
            ],
        });
      return config
    },
})
