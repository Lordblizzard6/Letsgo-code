import { mount } from "svelte";
import App from "./App.svelte";
import "./app.css";

const target = document.querySelector<HTMLElement>("#app") ?? document.body;

const app = mount(App, { target });

console.log("LetsGO UI initialized");