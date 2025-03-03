import React, { CSSProperties, useState } from "react";
import { Route, Routes } from "react-router-dom";
import { ThemeProvider } from "../hooks/theme-context";
import { Home } from "../pages";
import { Historical } from "../pages/historical";
import { OverUnder } from "../pages/overunder";
import { Appearance, Fueltype } from "../utils/types";
import { Footer } from "./footer";
import { Header } from "./header";

export const Layout: React.FC = () => {
  const [fueltype, setFueltype] = useState<Fueltype>("unleaded95");

  const searchParams = new URLSearchParams(window.location.search);
  const bgColor = searchParams.get("bgColor");
  const textColor = searchParams.get("textColor");
  let colorClasses = "";
  const inlineStyle: CSSProperties = {};
  if (!bgColor) {
    colorClasses += "bg-white dark:bg-slate-800 ";
  } else {
    inlineStyle.backgroundColor = `#${bgColor}`;
  }
  if (!textColor) {
    colorClasses += "text-slate-900 dark:text-white";
  } else {
    inlineStyle.color = `#${textColor}`;
  }
  const appearance: Appearance = { bgColor, textColor };
  return (
    <ThemeProvider>
      <div
        className={`h-screen flex flex-col ${colorClasses}`}
        style={inlineStyle}
      >
        <Header fueltype={fueltype} setFueltype={setFueltype} />
        <main className="mb-auto mt-4">
          <Routes>
            <Route
              path="/"
              element={<Home fueltype={fueltype} appearance={appearance} />}
            />
            <Route
              path="/historical"
              element={<Historical fueltype={fueltype} />}
            />
            <Route
              path="/overunder"
              element={<OverUnder fueltype={fueltype} />}
            />
          </Routes>
        </main>
        <Footer />
      </div>
    </ThemeProvider>
  );
};
