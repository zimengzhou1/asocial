import React, { useState, useEffect } from "react";

interface Props {
  style: React.CSSProperties;
  data: string;
  textKey: string;
}

const ExternalMessage: React.FC<Props> = ({ style, data, textKey }) => {
  return (
    <div id={textKey}>
      <p
        style={{
          ...style,
          position: "absolute",
          border: "none",
          background: "transparent",
          outline: "none",
          fontFamily: "Nunito, sans-serif",
          whiteSpace: "nowrap",
        }}
      >
        {data}
      </p>
    </div>
  );
};

export default ExternalMessage;
