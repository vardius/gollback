// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	site: 'https://vardius.github.io',
	base: '/gollback',
	trailingSlash: 'always',
	// Pages import the runnable programs in ../examples, so the dev server may read outside this folder.
	vite: { server: { fs: { allow: ['..'] } } },
	integrations: [
		starlight({
			title: 'gollback',
			description:
				'Race, All and Retry for Go: run functions concurrently and get typed results back, with generics and context cancellation.',
			logo: { src: './src/assets/mark.svg', alt: '' },
			favicon: '/favicon.svg',
			components: { Hero: './src/components/Hero.astro' },
			// Long lines wrap instead of scrolling under the copy button on phones.
			expressiveCode: { defaultProps: { wrap: true } },
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/vardius/gollback' }],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Start here',
					items: ['getting-started'],
				},
				{
					label: 'Functions',
					items: ['race', 'all', 'retry'],
				},
				{
					label: 'Reference',
					items: [
						'benchmarks',
						'migrating-from-v1',
						{ label: 'API reference', link: 'https://pkg.go.dev/github.com/vardius/gollback/v2' },
					],
				},
			],
		}),
	],
});
