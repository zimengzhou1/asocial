import React, { use, useEffect, useState, useRef } from "react";
import SideMenu from "@/components/SideMenu";
import LocalMessage from "@/components/LocalMessage";
import ExternalMessage from "@/components/ExternalMessage";
import Messages from "@/components/Messages";

import uniqid from "uniqid";
import {
  TransformWrapper,
  TransformComponent,
  useTransformEffect,
} from "react-zoom-pan-pinch";

const localuserID = uniqid();
const REMOVE_DELAY = 5000;

interface TextComponent {
  key: string;
  user: string;
  data: string;
  timeoutID: number;
  posX: number;
  posY: number;
  fadeOut: boolean;
}

interface Texts {
  [key: string]: TextComponent;
}

// interface user to list of messages
interface UserMessages {
  [key: string]: string[];
}

const ChatPage: React.FC = () => {
  return (
    <>
      <div className="fixed top-0 left-0 p-4 z-10">
        <SideMenu />
      </div>
      {/* <TransformWrapper
        pinch={{ step: 50 }}
        centerOnInit={true}
        initialScale={1}
        panning={{ velocityDisabled: true }}
        limitToBounds={true}
        doubleClick={{ disabled: true }}
        disablePadding={true}
      >
        <TransformComponent
          wrapperStyle={{
            background: "#f7f7f7",
            width: "100vw",
            height: "100vh",
          }}
        > */}
      <Messages />
      {/* </TransformComponent>
      </TransformWrapper> */}
    </>
  );
};

export default ChatPage;
