// next.config.js

module.exports = {
    pageExtensions: ['js', 'jsx', 'mdx'],

    webpack: (config) => {
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
}
