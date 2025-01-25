import React, { useState } from 'react';
import NamespaceSelector from './components/NamespaceSelector';
import DeploymentManager from './components/DeploymentManager';
import PodManager from './components/PodManager';
import ClusterInfo from './components/ClusterInfo';

const App: React.FC = () => {
  const [namespace, setNamespace] = useState<string>('');

  const handleNamespaceChange = (selectedNamespace: string) => {
    setNamespace(selectedNamespace);
  };

  return (
    <div>
      <h1>Cluster Manager</h1>

      <NamespaceSelector selectedNamespace={namespace} onNamespaceChange={handleNamespaceChange} />
      <ClusterInfo />

      {namespace && (
        <>
          <DeploymentManager namespace={namespace} />
          <PodManager namespace={namespace} />
        </>
      )}
    </div>
  );
};

export default App;
