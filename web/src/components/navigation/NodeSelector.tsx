import React from 'react';
import './node_selector.scss'
import {
  TextBulletListCheckmark20Regular as SelectNodeIcon
} from "@fluentui/react-icons";

function NodeSelector() {
  return (
      <div className="node-selector">
        <div className="content">

          <div className="title">
            <div className="text">Routing Node 2</div>
          </div>

          <div className="actions">
            <div className="icon"><SelectNodeIcon/></div>
          </div>

        </div>
      </div>
    );
}
export default NodeSelector;
