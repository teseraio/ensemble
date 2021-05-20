import React from 'react';

import Header from "./header"

export default {
  title: 'App/Header',
  component: Header,
};

const Template = (args) => <Header />

export const Primary = Template.bind({});
Primary.args = {
  primary: true,
  label: 'App',
};
