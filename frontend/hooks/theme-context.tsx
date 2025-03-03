import React from "react";

type ThemeType = "light" | "dark" | "auto";

function storageGet(): ThemeType {
  return (localStorage.getItem("theme") ?? "auto") as ThemeType;
}

function storageSet(theme: ThemeType) {
  localStorage.setItem("theme", theme);
}

function getPrefersDarkColorScheme() {
  const prefersDarkColorScheme = window.matchMedia(
    "(prefers-color-scheme: dark)"
  );
  return prefersDarkColorScheme;
}

function setDocumentThemeClass(prefersDarkColorScheme: MediaQueryList) {
  const theme = storageGet();
  switch (theme) {
    case "auto":
      {
        const darkOS = prefersDarkColorScheme.matches;
        if (darkOS) {
          document.documentElement.classList.add("dark");
        } else {
          document.documentElement.classList.remove("dark");
        }
      }
      break;
    case "dark":
      document.documentElement.classList.add("dark");
      break;
    case "light":
      document.documentElement.classList.remove("dark");
      break;
  }
}

interface ThemeContext {
  theme: ThemeType;
  toggle: () => void;
}

const defaultContextData: ThemeContext = {
  theme: "auto",
  toggle: () => {},
};

const themeContext = React.createContext(defaultContextData);
const useTheme = () => React.useContext(themeContext);

interface ThemeState {
  theme: ThemeType;
}
const useEffectDarkMode = (): [
  ThemeState,
  React.Dispatch<React.SetStateAction<ThemeState>>
] => {
  const [themeState, setThemeState] = React.useState<ThemeState>({
    theme: "auto",
  });

  React.useEffect(() => {
    const prefersDarkColorScheme = getPrefersDarkColorScheme();
    setDocumentThemeClass(prefersDarkColorScheme);
    setThemeState({ theme: storageGet() });
    const eventListenerHandler = () =>
      setDocumentThemeClass(prefersDarkColorScheme);
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
    let nextState: ThemeType = "auto";
    const prefersDarkColorScheme = getPrefersDarkColorScheme();
    switch (themeState.theme) {
      case "auto":
        {
          // If we are already in dark mode, go to light mode
          const darkOS = prefersDarkColorScheme.matches;
          if (darkOS) {
            nextState = "light";
          } else {
            nextState = "dark";
          }
        }
        break;
      case "dark":
        nextState = "light";
        break;
      case "light":
        nextState = "auto";
        break;
    }
    storageSet(nextState);
    setDocumentThemeClass(prefersDarkColorScheme);
    setThemeState({ theme: nextState });
  };

  return (
    <themeContext.Provider
      value={{
        theme: themeState.theme,
        toggle,
      }}
    >
      {children}
    </themeContext.Provider>
  );
};

// eslint-disable-next-line react-refresh/only-export-components
export { ThemeProvider, useTheme };
