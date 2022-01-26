defmodule ControlServerWeb.ServicesLive.Postgres.FormComponent do
  use ControlServerWeb, :live_component

  alias ControlServer.Postgres
  alias ControlServer.Postgres.Cluster

  require Logger

  @impl true
  def mount(socket) do
    {:ok,
     socket
     |> assign_new(:save_info, fn -> "cluster:save" end)
     |> assign_new(:save_target, fn -> nil end)}
  end

  @impl true
  def update(%{cluster: cluster} = assigns, socket) do
    changeset = Postgres.change_cluster(cluster)

    {:ok,
     socket
     |> assign(assigns)
     |> assign(:changeset, changeset)}
  end

  @impl true
  def handle_event("validate", %{"cluster" => params}, socket) do
    {changeset, data} = Cluster.validate(params)
    {:noreply, assign(socket, changeset: changeset, data: data)}
  end

  def handle_event("save", %{"cluster" => cluster_params}, socket) do
    save_cluster(socket, socket.assigns.action, cluster_params)
  end

  defp save_cluster(socket, :new, cluster_params) do
    case Postgres.create_cluster(cluster_params) do
      {:ok, new_cluster} ->
        {:noreply,
         socket
         |> put_flash(:info, "Postgres Cluster created successfully")
         |> send_info(socket.assigns.save_target, new_cluster)}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, changeset: changeset)}
    end
  end

  defp save_cluster(socket, :edit, cluster_params) do
    case Postgres.update_cluster(socket.assigns.cluster, cluster_params) do
      {:ok, updated_cluster} ->
        {:noreply,
         socket
         |> put_flash(:info, "Postgres Cluster updated successfully")
         |> send_info(socket.assigns.save_target, updated_cluster)}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, :changeset, changeset)}
    end
  end

  defp send_info(socket, nil, _cluster), do: {:noreply, socket}

  defp send_info(socket, target, cluster) do
    send(target, {socket.assigns.save_info, %{"cluster" => cluster}})
    socket
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="space-y-10">
      <.form
        let={f}
        for={@changeset}
        id="cluster-form"
        phx-change="validate"
        phx-submit="save"
        phx-target={@myself}
      >
        <div class="grid grid-cols-1 mt-6 gap-y-6 gap-x-4 sm:grid-cols-6">
          <.form_field
            type="text_input"
            form={f}
            field={:name}
            placeholder="Name"
            wrapper_class="sm:col-span-3"
          />
          <.form_field
            type="number_input"
            form={f}
            field={:num_instances}
            placeholder="Number of Instances"
            wrapper_class="sm:col-span-3"
          />
          <.form_field
            type="text_input"
            form={f}
            field={:postgres_version}
            placeholder="Postgres Version"
            wrapper_class="sm:col-span-3"
          />
          <.form_field
            type="text_input"
            form={f}
            field={:size}
            placeholder="Size"
            wrapper_class="sm:col-span-3"
          />
          <.button type="submit" phx_disable_with="Saving…">
            Save
          </.button>
        </div>
      </.form>
    </div>
    """
  end
end
