import { Metadata } from "next";
import { notFound } from "next/navigation";
import { DictionaryAPI } from "@/lib/dictionary-api";
import WordDetail from "@/components/dictionary/WordDetail";
import SearchBar from "@/components/dictionary/SearchBar";
import Link from "next/link";

interface PageProps {
  params: Promise<{ word: string }>;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { word } = await params;
  const decoded = decodeURIComponent(word);
  return {
    title: `${decoded} - Juliy Dictionary`,
    description: `Meaning, pronunciation, and examples for the word ${decoded}.`,
    alternates: {
      canonical: `https://juliy.com/dictionary/${encodeURIComponent(word)}`,
    }
  };
}

export default async function WordPage({ params }: PageProps) {
  const { word } = await params;
  const decodedWord = decodeURIComponent(word);
  const data = await DictionaryAPI.getDetail(decodedWord);

  if (!data) {
    return (
      <div className="min-h-[calc(100vh-4rem)] pt-12 px-6">
        <div className="max-w-3xl mx-auto mb-10">
          <SearchBar />
        </div>
        <div className="text-center py-20">
          <h1 className="text-4xl font-bold text-white mb-4">Word Not Found</h1>
          <p className="text-white/60 mb-8">
            Sorry, we couldn't find the word <span className="font-semibold text-white">"{decodedWord}"</span> in our dictionary.
          </p>
          <Link
            href="/"
            className="px-6 py-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors text-white font-medium"
          >
            Go back home
          </Link>
        </div>
      </div>
    );
  }

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "DefinedTerm",
    "name": data.word,
    "description": data.definitions?.[0]?.definition || `Definition of ${data.word}`,
    "inDefinedTermSet": "https://juliy.com/dictionary",
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] pt-12 px-6 pb-20">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <div className="max-w-3xl mx-auto mb-12">
        <SearchBar />
      </div>
      <WordDetail word={data} />
    </div>
  );
}
