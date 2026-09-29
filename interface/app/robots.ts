import { MetadataRoute } from 'next'

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: ['/profile/', '/api/', '/_next/', '/placement/', '/match/', '/learn/'],
    },
    sitemap: 'https://juliyenglish.id.vn/sitemap.xml',
  }
}
