import React, { useState, useEffect } from 'react';
import NamespaceSelector from './components/NamespaceSelector';
import DeploymentManager from './components/DeploymentManager';
import PodManager from './components/PodManager';
import ClusterInfo from './components/ClusterInfo';

const App: React.FC = () => {
  const [namespace, setNamespace] = useState<string>('');
  const [darkMode, setDarkMode] = useState<boolean>(false);

  const handleNamespaceChange = (selectedNamespace: string) => {
    setNamespace(selectedNamespace);
  };

  const toggleTheme = () => {
    setDarkMode((prev) => !prev);
  };

  useEffect(() => {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
      setDarkMode(savedTheme === 'dark');
      document.documentElement.classList.toggle('dark', savedTheme === 'dark');
    }
  }, []);

  useEffect(() => {
    const theme = darkMode ? 'dark' : 'light';
    localStorage.setItem('theme', theme);
    document.documentElement.classList.toggle('dark', darkMode);
  }, [darkMode]);

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors">
      <header
        className={`${darkMode ? 'bg-blue-900' : 'bg-blue-600'
          } text-white py-4 shadow-lg transition-colors`}
      >
        <div className="max-w-6xl mx-auto px-4 flex justify-between items-center">
          <h1 className="text-2xl font-semibold">Cluster Manager</h1>
          <button
            onClick={toggleTheme}
            className={`${darkMode ? 'bg-blue-600' : 'bg-gray-800'
              } text-white px-4 py-2 rounded-md transition-transform transform hover:scale-105`}
          >
            {darkMode ? 'Light Mode' : 'Dark Mode'}
          </button>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="flex flex-col gap-6">
          <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 transition-transform hover:scale-105">
            <h2 className="text-lg font-semibold mb-4">Select Namespace</h2>
            <NamespaceSelector
              selectedNamespace={namespace}
              onNamespaceChange={handleNamespaceChange}
            />
          </div>
          <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 transition-transform hover:scale-105">
            <ClusterInfo />
          </div>
          {namespace && (
            <div className="grid gap-6 lg:grid-cols-2 xl:grid-cols-2">
              <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 transition-transform hover:scale-105">
                <DeploymentManager namespace={namespace}/>
              </div>
              <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 transition-transform hover:scale-105">
                <PodManager namespace={namespace}/>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default App;
