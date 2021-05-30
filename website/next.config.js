// next.config.js

const withTM = require('next-transpile-modules')(['@teseraio/tesera-oss', '@teseraio/oss-react-changelog', '@teseraio/oss-react-docs', '@teseraio/oss-react-landing', '@teseraio/oss-react-app', '@teseraio/oss-react-community', '@teseraio/cookie-consent-manager']);

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
