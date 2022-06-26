import { useTheme } from "../hooks/theme-context";

export const Footer: React.FC = () => {
  const { theme, toggle } = useTheme();
  return (
    <footer className="p-2">
      <div className="flex flex-row justify-around text-sm">
        <div>
          Data source:{" "}
          <a className="underline" href="https://ok.dk">
            ok.dk
          </a>
        </div>
        <button title="Toggle color theme" onClick={() => toggle()}>
          {theme === "auto" && "🌓"}
          {theme === "dark" && "🌚"}
          {theme === "light" && "🌞"}
        </button>
      </div>
    </footer>
  );
};
