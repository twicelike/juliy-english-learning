import { DictionaryAPI } from "@/lib/dictionary-api";
import SearchBar from "@/components/dictionary/SearchBar";
import WordCard from "@/components/dictionary/WordCard";
import RandomWordCard from "@/components/dictionary/RandomWordCard";

// Server Component
export default async function HomePage() {
  let wordOfTheDay = null;
  let randomWord = null;

  try {
    // We can fetch these server-side for fast initial load
    [wordOfTheDay, randomWord] = await Promise.all([
      DictionaryAPI.getWordOfTheDay().catch(() => null),
      DictionaryAPI.getRandom().catch(() => null)
    ]);
  } catch (err) {
    // Silently handle errors on the landing page
  }

  return (
    <div className="relative flex flex-col items-center justify-start min-h-[calc(100vh-4rem)] pt-20 px-6 overflow-hidden">
      {/* Background gradients */}
      <div className="pointer-events-none absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] rounded-[100%] bg-indigo-600/10 blur-[120px]" />

      {/* Hero content */}
      <div className="relative z-10 w-full max-w-4xl text-center mb-16">

        <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-white mb-6 leading-tight">
          Dictionary & <br />
          <span className="bg-gradient-to-r from-indigo-400 to-violet-400 bg-clip-text text-transparent">
            AI Tutor
          </span>
        </h1>

        {/* Search Bar */}
        <div className="mb-12">
          <SearchBar />
        </div>
      </div>

      {/* Daily features */}
      <div className="relative z-10 w-full max-w-4xl grid grid-cols-1 md:grid-cols-2 gap-6 pb-20">
        {wordOfTheDay && (
          <WordCard word={wordOfTheDay} title="Word of the Day" />
        )}
        {randomWord && (
          <RandomWordCard initialWord={randomWord} />
        )}
      </div>
    </div>
  );
}
