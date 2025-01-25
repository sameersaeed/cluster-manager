import React, { useEffect, useState } from 'react';
import axios from 'axios';

interface Node {
    name: string;
    status: any;
    roles: string;
    cpu: string;
    memory: string;
}

const NodeDetails: React.FC = () => {
    const [nodes, setNodeDetails] = useState<Node[]>([]);

    useEffect(() => {
        axios.get("http://localhost:8080/api/node-details")
          .then((response) => {
            console.log(response.data);
            setNodeDetails(response.data.nodes); 
          })
          .catch((error) => {
            console.error("Error fetching node details:", error);
          });
      }, []);
    return (
        <div>
            <h2>Node Details</h2>
            <ul>
                {nodes.map((node) => (
                    <li key={node.name}>
                        <strong>{node.name}</strong> (Status: {node.status}) <br />
                        CPU: {node.cpu}, Memory: {node.memory}
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default NodeDetails;
