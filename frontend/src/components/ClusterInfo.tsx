import React, { useEffect, useState } from 'react';
import axios from 'axios';

const ClusterInfo: React.FC = () => {
  const [clusterName, setClusterName] = useState<string>('');

  useEffect(() => {
    axios.get('http://localhost:8080/api/cluster-name')
      .then(response => setClusterName(response.data.clusterName))
      .catch(error => console.error("Error fetching cluster name:", error));
  }, []);

  return (
    <div>
      <h3>Cluster Name: {clusterName}</h3>
    </div>
  );
};

export default ClusterInfo;
