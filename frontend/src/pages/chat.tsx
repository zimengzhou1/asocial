import React, { use, useEffect, useState } from "react";
import SideMenu from "@/components/SideMenu";
import LocalMessage from "@/components/LocalMessage";
import uniqid from "uniqid";
import ExternalMessage from "@/components/ExternalMessage";

const localuserID = uniqid();
const REMOVE_DELAY = 5000;

interface TextComponent {
  key: string;
  user: string;
  data: string;
  timeoutID: number;
  posX: number;
  posY: number;
}

interface Texts {
  [key: string]: TextComponent;
}

// interface user to list of messages
interface UserMessages {
  [key: string]: string[];
}

const ChatPage: React.FC = () => {
  const [texts, setTexts] = useState<Texts>({});
  // Each user can show max 3 messages at a time
  const [userMessages, setUserMessages] = useState<UserMessages>({});
  const [socket, setSocket] = useState<WebSocket | null>(null);

  // Create socket connection
  useEffect(() => {
    console.log("Connecting to server");
    const socket = new WebSocket("ws://localhost:5001/api/chat");
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

  // useEffect(() => {
  //   console.log("User messages:", userMessages);
  // }, [userMessages]);

  // const handleIncomingMessage = (data: any) => {
  //   console.log("Incoming message:", data);
  //   const { userID, textID, data: textData, pos } = data;
  //   if (!(textID in texts)) {
  //     console.log("in here cunt");
  //     const timeoutID = window.setTimeout(
  //       removeInactiveComponents,
  //       REMOVE_DELAY,
  //       textID,
  //       userID
  //     );
  //     const newComponent: TextComponent = {
  //       key: textID,
  //       user: userID,
  //       data: textData,
  //       timeoutID: timeoutID,
  //       posX: pos.x,
  //       posY: pos.y,
  //     };
  //     console.log("timeoutID", timeoutID);
  //     // setTexts({ ...texts, [textID]: newComponent });
  //     setTexts((prevTexts) => {
  //       // Collect all the changes and perform a single state update
  //       return { ...prevTexts, [textID]: newComponent };
  //     });
  //     updateUserMessages(userID, textID);
  //   } else {
  //     console.log("already created!");
  //     handleInputChange(textID, textData);
  //   }
  // };

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
      delete updatedTexts[id];
      return updatedTexts;
    });
    setUserMessages((prevUserMessages) => {
      const updatedUserMessages = { ...prevUserMessages };
      console.log(updatedUserMessages);
      console.log(userID);
      updatedUserMessages[userID] = updatedUserMessages[userID].filter(
        (key) => key !== id
      );
      if (updatedUserMessages[userID].length === 0) {
        delete updatedUserMessages[userID];
      }
      return updatedUserMessages;
    });
  };

  const addLocalMessage = (event: { clientX: any; clientY: any }) => {
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
      posX: event.clientX,
      posY: event.clientY,
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
    <>
      <div className="absolute top-0 left-0 p-4 z-10">
        <SideMenu />
      </div>
      <div
        onClick={addLocalMessage}
        style={{ width: "100vw", height: "100vh", position: "relative" }}
      >
        {Object.values(texts).map((data, index) => (
          <React.Fragment key={index}>
            {data.user === localuserID ? (
              <LocalMessage
                key={data.timeoutID}
                style={{
                  top: `${data.posY}px`,
                  left: `${data.posX}px`,
                }}
                data={data.data}
                timeoutID={data.timeoutID}
                textKey={data.key}
                onInputChange={handleInputChange}
              />
            ) : (
              <ExternalMessage
                key={data.timeoutID}
                style={{
                  top: `${data.posY}px`,
                  left: `${data.posX}px`,
                }}
                data={data.data}
                textKey={data.key}
              />
            )}
          </React.Fragment>
        ))}
      </div>
    </>
  );
};

export default ChatPage;
