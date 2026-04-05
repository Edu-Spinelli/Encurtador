export default function TechBadge({ name }: { name: string }) {
  return (
    <span className="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium bg-zinc-800/80 text-zinc-400 border border-zinc-700/50">
      {name}
    </span>
  );
}
