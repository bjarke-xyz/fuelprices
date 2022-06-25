import React from "react";

interface ThemeContext {
  dark: boolean;
  toggle: () => void;
}

const defaultContextData: ThemeContext = {
  dark: false,
  toggle: () => {},
};

const themeContext = React.createContext(defaultContextData);
const useTheme = () => React.useContext(themeContext);

interface ThemeState {
  dark: boolean;
}
const useEffectDarkMode = (): [
  ThemeState,
  React.Dispatch<React.SetStateAction<ThemeState>>
] => {
  const [themeState, setThemeState] = React.useState<ThemeState>({
    dark: false,
  });

  React.useEffect(() => {
    function setTheme(prefersDarkColorScheme: MediaQueryList) {
      const darkOS = prefersDarkColorScheme.matches;
      if (darkOS) {
        document.documentElement.classList.add("dark");
      } else {
        document.documentElement.classList.remove("dark");
      }
      setThemeState({ ...themeState, dark: darkOS });
    }
    const prefersDarkColorScheme = window.matchMedia(
      "(prefers-color-scheme: dark)"
    );
    setTheme(prefersDarkColorScheme);
    const eventListenerHandler = () => setTheme(prefersDarkColorScheme);
    prefersDarkColorScheme.addEventListener("change", eventListenerHandler);

    return function cleanup() {
      prefersDarkColorScheme.removeEventListener(
        "change",
        eventListenerHandler
      );
    };
  }, []);

  return [themeState, setThemeState];
};

const ThemeProvider = ({ children }: { children: React.ReactNode }) => {
  const [themeState, setThemeState] = useEffectDarkMode();

  const toggle = () => {
    const dark = !themeState.dark;
    if (dark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
    setThemeState({ dark });
  };

  return (
    <themeContext.Provider
      value={{
        dark: themeState.dark,
        toggle,
      }}
    >
      {children}
    </themeContext.Provider>
  );
};

export { ThemeProvider, useTheme };
