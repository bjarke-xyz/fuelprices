import { Head, Html, Main, NextScript } from "next/document";

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <title>Fuel Prices</title>
        <meta name="viewport" content="width=device-width, user-scalable=no" />
      </Head>
      <body className="bg-white dark:bg-slate-800 text-slate-900 dark:text-white">
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}
