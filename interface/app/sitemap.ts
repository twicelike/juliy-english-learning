import { MetadataRoute } from 'next'

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    {
      url: 'https://juliyenglish.id.vn',
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 1,
    },
    {
      url: 'https://juliyenglish.id.vn/explore',
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.9,
    },
    {
      url: 'https://juliyenglish.id.vn/word-sets',
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.8,
    }
    // Note: For a dictionary with 100k+ words, Next.js allows paginated sitemaps using generateSitemaps()
    // For now, these are the core static routes.
  ]
}
