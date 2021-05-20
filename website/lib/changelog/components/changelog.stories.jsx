import React from 'react';

import Changelog from "./changelog"

export default {
  title: 'Changelog/Changelog',
  component: Changelog,
};

const Template = (args) => <Changelog {...args} />

const vers = [
    {
        version: "1.0",
        content: "xx"
    },
    {
        version: "1.1",
        content: "xx"
    }
]

export const Primary = Template.bind({});
Primary.args = {
  primary: true,
  label: 'Changelog',
  vers: vers,
};
