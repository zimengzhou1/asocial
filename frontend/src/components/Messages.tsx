import React, { use, useEffect, useState, useRef } from "react";
import SideMenu from "@/components/SideMenu";
import LocalMessage from "@/components/LocalMessage";
import ExternalMessage from "@/components/ExternalMessage";

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

const Messages: React.FC = ({}) => {
  const [texts, setTexts] = useState<Texts>({});
  // Each user can show max 3 messages at a time
  const [userMessages, setUserMessages] = useState<UserMessages>({});
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const messagesRef = useRef<HTMLDivElement>(null);

  // const [isDragging, setIsDragging] = useState(false);
  // // positionX and positionY are the current position of the transform
  // const [positions, setPositions] = useState({ x: 0, y: 0, scale: 1 });

  // useTransformEffect(({ state, instance }) => {
  //   setPositions({
  //     x: state.positionX,
  //     y: state.positionY,
  //     scale: state.scale,
  //   });

  //   return () => {
  //     // unmount
  //   };
  // });

  // const handleClick = (event: { clientX: any; clientY: any }) => {
  //   if (!isDragging) {
  //     console.log(positions.x, positions.y);
  //     const adjustedX = event.clientX / positions.scale - positions.x;
  //     const adjustedY = event.clientY / positions.scale - positions.y;
  //     console.log(adjustedX, adjustedY);
  //     addLocalMessage(adjustedX, adjustedY);
  //   }
  // };

  // Create socket connection
  useEffect(() => {
    // scroll into view
    if (messagesRef.current) {
      messagesRef.current.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "center",
      });
    }
    console.log("Connecting to server");
    const socket = new WebSocket("ws://" + window.location.host + "/api/chat");
    setSocket(socket);

    socket.onopen = () => {
      console.log("Connected to server");
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleIncomingMessage(data);
    };

    return () => {
      // Clean up the socket connection when the component unmounts
      socket.close();
    };
  }, []);

  const handleIncomingMessage = (data: any) => {
    const { userID, textID, data: textData, pos } = data;

    const timeoutID = window.setTimeout(
      removeInactiveComponents,
      REMOVE_DELAY,
      textID,
      userID
    );

    setTexts((prevTexts) => {
      if (!(textID in prevTexts)) {
        const newComponent: TextComponent = {
          key: textID,
          user: userID,
          data: textData,
          timeoutID: timeoutID,
          posX: pos.x,
          posY: pos.y,
          fadeOut: false,
        };

        updateUserMessages(userID, textID);
        return { ...prevTexts, [textID]: newComponent };
      } else {
        const oldTimeoutID = prevTexts[textID].timeoutID;
        window.clearTimeout(oldTimeoutID);
        return {
          ...prevTexts,
          [textID]: {
            ...prevTexts[textID],
            timeoutID: timeoutID,
            data: textData,
          },
        };
      }
    });
  };

  const removeInactiveComponents = (id: string, userID: string) => {
    //console.log("removing...");
    setTexts((prevTexts) => {
      // if remove triggered by too many messages, remove timeout
      window.clearTimeout(prevTexts[id].timeoutID);

      const updatedTexts = { ...prevTexts };
      updatedTexts[id] = { ...updatedTexts[id], fadeOut: true };

      setTimeout(() => {
        setTexts((prevTexts) => {
          const updatedTexts = { ...prevTexts };
          delete updatedTexts[id];
          return updatedTexts;
        });
      }, 500);
      return updatedTexts;
    });
    setUserMessages((prevUserMessages) => {
      const updatedUserMessages = { ...prevUserMessages };
      // console.log(updatedUserMessages);
      // console.log(userID);
      updatedUserMessages[userID] = updatedUserMessages[userID].filter(
        (key) => key !== id
      );
      if (updatedUserMessages[userID].length === 0) {
        delete updatedUserMessages[userID];
      }
      return updatedUserMessages;
    });
  };

  const addLocalMessage = (event: { pageX: any; pageY: any }) => {
    const posX = event.pageX;
    const posY = event.pageY;
    const newKey = uniqid();
    const newData = "";
    const timeoutID = window.setTimeout(
      removeInactiveComponents,
      REMOVE_DELAY,
      newKey,
      localuserID
    );
    const newComponent: TextComponent = {
      key: newKey,
      user: localuserID,
      data: newData,
      timeoutID: timeoutID,
      posX: posX,
      posY: posY,
      fadeOut: false,
    };

    setTexts({ ...texts, [newKey]: newComponent });
    updateUserMessages(localuserID, newKey);
  };

  const updateUserMessages = (userID: string, textKey: string) => {
    setUserMessages((prevUserMessages) => {
      const updatedUserMessages = { ...prevUserMessages };
      let newMessages: string[] = [];
      if (updatedUserMessages[userID] !== undefined) {
        newMessages = [...updatedUserMessages[userID]];
      }

      if (newMessages.length > 4) {
        const oldKey = newMessages.shift();
        if (oldKey) {
          removeInactiveComponents(oldKey, userID);
        }
      }

      if (!newMessages.includes(textKey)) {
        newMessages.push(textKey);
      }

      return { ...prevUserMessages, [userID]: newMessages };
    });
  };

  const handleInputChange = (textKey: string, data: string) => {
    // Clear and create new timeout
    const timeoutID = texts[textKey].timeoutID;
    window.clearTimeout(timeoutID);
    const newTimeoutID = window.setTimeout(
      removeInactiveComponents,
      REMOVE_DELAY,
      textKey,
      texts[textKey].user
    );

    // Send data to server
    if (socket && texts[textKey].user === localuserID) {
      socket.send(
        JSON.stringify({
          userID: texts[textKey].user,
          textID: textKey,
          data: data,
          pos: { x: texts[textKey].posX, y: texts[textKey].posY },
        })
      );
    }

    setTexts((prevTexts) => {
      return {
        ...prevTexts,
        [textKey]: {
          ...prevTexts[textKey],
          timeoutID: newTimeoutID,
          data: data,
        },
      };
    });
  };

  return (
    <div
      onClick={addLocalMessage}
      style={{
        width: "100vw",
        height: "100vh",
        position: "relative",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        background: "#f7f7f7",
      }}
    >
      <p
        ref={messagesRef}
        style={{
          fontFamily: "Nunito, sans-serif",
          fontSize: "0.875rem",
        }}
      >
        click anywhere to type
      </p>
      {Object.values(texts).map((data, index) => (
        <React.Fragment key={index}>
          {data.user === localuserID ? (
            <LocalMessage
              key={data.timeoutID}
              style={{
                top: `${data.posY}px`,
                left: `${data.posX}px`,
                position: "fixed",
              }}
              data={data.data}
              timeoutID={data.timeoutID}
              textKey={data.key}
              fadeOut={data.fadeOut}
              onInputChange={handleInputChange}
            />
          ) : (
            <ExternalMessage
              key={data.timeoutID}
              style={{
                top: `${data.posY}px`,
                left: `${data.posX}px`,
                position: "fixed",
              }}
              data={data.data}
              textKey={data.key}
              fadeOut={data.fadeOut}
            />
          )}
        </React.Fragment>
      ))}
    </div>
  );
};

export default Messages;
