import SideMenu from "@/components/SideMenu";
import Link from "next/link";

export default function Home() {
  return (
    <div className="flex flex-col justify-center items-center h-screen">
      <div className="absolute top-0 left-0 p-4">
        <SideMenu />
      </div>
      <h1 className="text-4xl font-custom mb-4">asocialpage</h1>
      <p
        className="font-custom mb-16 relative w-[max-content]
before:absolute before:inset-0 before:animate-typewriter
before:bg-white
after:absolute after:inset-0 after:w-[0.125em] after:animate-caret
after:bg-black"
      >
        Talk with a keyboard
        <br />
      </p>
      <Link href="/chat" className="font-custom text-blue-600">
        start chatting
      </Link>
    </div>
  );
}
