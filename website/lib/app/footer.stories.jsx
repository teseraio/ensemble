import React from 'react';

import Footer from "./footer"

export default {
  title: 'App/Footer',
  component: Footer,
};

const Template = (args) => <Footer />

export const Primary = Template.bind({});
Primary.args = {
  primary: true,
  label: 'App',
};
