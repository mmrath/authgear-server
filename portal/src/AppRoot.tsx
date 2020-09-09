import React, { useMemo } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";
import { makeClient } from "./graphql/adminapi/apollo";
import ScreenLayout from "./ScreenLayout";
import UsersScreen from "./graphql/adminapi/UsersScreen";
import AuthenticationConfigurationScreen from "./graphql/portal/AuthenticationConfigurationScreen";

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams();
  const client = useMemo(() => {
    return makeClient(appID);
  }, [appID]);
  return (
    <ApolloProvider client={client}>
      <ScreenLayout>
        <Routes>
          <Route path="/" element={<Navigate to="users" replace={true} />} />
          <Route path="/users" element={<UsersScreen />} />
          <Route
            path="/configuration/authentication"
            element={<AuthenticationConfigurationScreen />}
          />
        </Routes>
      </ScreenLayout>
    </ApolloProvider>
  );
};

export default AppRoot;
