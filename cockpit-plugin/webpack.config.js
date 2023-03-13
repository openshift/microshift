import fs from "fs";

import copy from "copy-webpack-plugin";
import extract from "mini-css-extract-plugin";
import TerserJSPlugin from 'terser-webpack-plugin';
import CssMinimizerPlugin from 'css-minimizer-webpack-plugin';
import CompressionPlugin from "compression-webpack-plugin";
import ESLintPlugin from 'eslint-webpack-plugin';
import StylelintPlugin from 'stylelint-webpack-plugin';

import CockpitPoPlugin from "./pkg/lib/cockpit-po-plugin.js";
import CockpitRsyncPlugin from "./pkg/lib/cockpit-rsync-plugin.js";

/* A standard nodejs and webpack pattern */
const production = process.env.NODE_ENV === 'production';

/* development options for faster iteration */
const eslint = process.env.ESLINT !== '0';

/* Default to disable csslint for faster production builds */
const stylelint = process.env.STYLELINT ? (process.env.STYLELINT !== '0') : !production;

// Obtain package name from package.json
const packageJson = JSON.parse(fs.readFileSync('package.json'));

// Non-JS files which are copied verbatim to dist/
const copy_files = [
    "./src/index.html",
    "./src/manifest.json",
];

const plugins = [
    new copy({ patterns: copy_files }),
    new extract({ filename: "[name].css" }),
    new CockpitPoPlugin(),
    new CockpitRsyncPlugin({ dest: packageJson.name }),
];

if (eslint) {
    plugins.push(new ESLintPlugin({ extensions: ["js", "jsx"], failOnWarning: true, }));
}

if (stylelint) {
    plugins.push(new StylelintPlugin({
        context: "src/",
    }));
}

/* Only minimize when in production mode */
if (production) {
    plugins.unshift(new CompressionPlugin({
        test: /\.(js|html|css)$/,
        deleteOriginalAssets: true
    }));
}

const config = {
    mode: production ? 'production' : 'development',
    resolve: {
        modules: ["node_modules", 'pkg/lib'],
        alias: { 'font-awesome': 'font-awesome-sass/assets/stylesheets' },
    },
    resolveLoader: {
        modules: ["node_modules", 'pkg/lib'],
    },
    watchOptions: {
        ignored: /node_modules/,
    },
    entry: {
        index: "./src/index.js",
    },
    // cockpit.js gets included via <script>, everything else should be bundled
    externals: { cockpit: "cockpit" },
    devtool: "source-map",
    stats: "errors-warnings",
    // always regenerate dist/, so make rules work
    output: { clean: true, compareBeforeEmit: false },

    optimization: {
        minimize: production,
        minimizer: [
            new TerserJSPlugin({
                extractComments: {
                    condition: true,
                    filename: `[file].LICENSE.txt?query=[query]&filebase=[base]`,
                    banner(licenseFile) {
                        return `License information can be found in ${licenseFile}`;
                    },
                },
            }),
            new CssMinimizerPlugin()
        ],
    },

    module: {
        rules: [
            {
                exclude: /node_modules/,
                use: "babel-loader",
                test: /\.(js|jsx)$/
            },
            /* HACK: remove unwanted fonts from PatternFly's css */
            {
                test: /patternfly-4-cockpit.scss$/,
                use: [
                    extract.loader,
                    {
                        loader: 'css-loader',
                        options: {
                            sourceMap: true,
                            url: false,
                        },
                    },
                    {
                        loader: 'string-replace-loader',
                        options: {
                            multiple: [
                                {
                                    search: /src:url\("patternfly-icons-fake-path\/pficon[^}]*/g,
                                    replace: 'src:url("../base1/fonts/patternfly.woff") format("woff");',
                                },
                                {
                                    search: /@font-face[^}]*patternfly-fonts-fake-path[^}]*}/g,
                                    replace: '',
                                },
                            ]
                        },
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: !production,
                            sassOptions: {
                                outputStyle: production ? 'compressed' : undefined,
                            },
                        },
                    },
                ]
            },
            {
                test: /\.s?css$/,
                exclude: /patternfly-4-cockpit.scss/,
                use: [
                    extract.loader,
                    {
                        loader: 'css-loader',
                        options: {
                            sourceMap: true,
                            url: false
                        }
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: !production,
                            sassOptions: {
                                outputStyle: production ? 'compressed' : undefined,
                            },
                        },
                    },
                ]
            },
        ]
    },
    plugins
};

export default config;
