import { type NextPage } from "next";
import Head from "next/head";
import { api } from "~/utils/api";
import JsonEditor from "react-json-editor-ui";
import { useState } from "react";
import { BreaseProvider } from "@brease/react";
import { PropsWithChildren } from "react";

import "react-json-editor-ui/dist/react-json-editor-ui.cjs.development.css";

const Example = () => {
  const [editObject, setEditObject] = useState<any>({
    user_id: 1,
    first_name: "Gusella",
    last_name: "Dakers",
    email: "gdakers0@va.gov",
    age: 98,
    gender: "Female",
    address: "248 Parkside Hill",
    city: "Komysh-Zorya",
    state: null,
    country: "Ukraine",
    phone_number: "419-767-5757",
    job_title: "Financial Advisor",
    company_name: "Livefish",
    favorite_color: ["green", "red"],
    birthdate: "11/12/2021",
  });

  return (
    <JsonEditor
      data={editObject}
      onChange={(data) => {
        setEditObject(data);
      }}
      optionsMap={{
        color: [
          { value: "red", label: "Red" },
          { value: "blue", label: "Blue" },
        ],
        city: [
          { value: "beijing", label: "Beijing" },
          { value: "shanghai", label: "Shanghai" },
        ],
      }}
    />
  );
};

const BreaseContext = ({ children }: PropsWithChildren<{}>) => {
  const { data: accessToken, isLoading } =
    api.example.breaseWebToken.useQuery();
  if (isLoading || !accessToken) {
    return null;
  }

  return (
    <BreaseProvider
      accessToken={accessToken.accessToken}
      refreshToken={accessToken.refreshToken}
    >
      {children}
    </BreaseProvider>
  );
};

const Home: NextPage = () => {
  return (
    <>
      <Head>
        <title>brease NextJS example</title>
        <meta name="description" content="brease NextJS example" />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <main className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-b from-[#066d02] to-[#152c19]">
        <div className="container flex flex-col items-center justify-center gap-12 px-4 py-16 ">
          <h1 className="text-5xl font-extrabold tracking-tight text-white sm:text-[5rem]">
            brease<span className="text-[hsl(100,100%,70%)]">.run</span>
          </h1>
          <BreaseContext>
            <Example />
          </BreaseContext>
        </div>
      </main>
    </>
  );
};

export default Home;
